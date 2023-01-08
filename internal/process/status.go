package process

import (
	"context"
	"firebase.google.com/go/messaging"
	"fmt"
	"home-reminder-push/internal/clients/darksky"
	"home-reminder-push/internal/clients/firebase"
	"home-reminder-push/internal/models"
	"log"
	"time"
)

type Status struct {
	darksky  *darksky.Client
	firebase *firebase.Client
}

func New(darkskyClient *darksky.Client, firebaseClient *firebase.Client) *Status {
	return &Status{
		darkskyClient,
		firebaseClient,
	}
}

func (s *Status) Start(done <-chan struct{}) <-chan error {
	ch := make(chan error)

	dayTimer := time.NewTicker(time.Second)
	hourlyTimer := time.Tick(time.Hour)

	var sendHourlyWarningAt int64 = 0

	go func() {
		for {
			select {
			case <-done:
				log.Println("stopping status loop")
				dayTimer.Stop()
				close(ch)
				return
			case <-dayTimer.C:
				dayTimer.Reset(getDurationTillTomorrow())
				log.Println("getting daily forecast")
				f, err := s.darksky.GetForecast()
				if err != nil {
					log.Println(err)
				}

				log.Println("sending daily message")
				warn, timeAt := isCold(f)
				err = s.sendDailyMessage(warn, f.Daily.Data[0].TemperatureLow)
				if err != nil {
					log.Println(err)
				}
				sendHourlyWarningAt = timeAt
			case <-hourlyTimer:
				if sendHourlyWarningAt == 0 {
					log.Println("no hourly warning, doing nothing")
				} else if time.UnixMilli(sendHourlyWarningAt).Sub(time.Now()) <= time.Hour*3 || time.Now().Hour() > 21 {
					err := s.sendWarningMessage(sendHourlyWarningAt)
					if err != nil {
						log.Println(err)
					}
					sendHourlyWarningAt = 0
				}
			}
		}
	}()

	return ch
}

func (s *Status) sendDailyMessage(warn bool, tempLow float64) error {
	var (
		title = ""
		body  = ""
	)
	if warn {
		title = "Bring the plants in later!"
		body = fmt.Sprintf(
			"Looks like the temperature for today will be cold with a low of %.0f, consider bringing the plants in!",
			tempLow,
		)
	} else {
		title = "No need to bring the plants in today :)"
		body = fmt.Sprintf(
			"The coldest it's going to reach is %.0f.",
			tempLow,
		)
	}
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Topic: "home-reminder",
	}
	_, err := s.firebase.Send(context.Background(), message)
	if err != nil {
		return fmt.Errorf("error sending firebase message: %w", err)
	}

	return nil
}

func (s *Status) sendWarningMessage(t int64) error {
	var (
		title = ""
		body  = ""
	)

	formattedTime := time.UnixMilli(t).Format("15:04:05 pm")
	title = "Remember to bring the plants in, it's going to be cold soon!"
	body = fmt.Sprintf("Bring the plants in before %s!", formattedTime)
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Topic: "home-reminder",
	}
	_, err := s.firebase.Send(context.Background(), message)
	if err != nil {
		return fmt.Errorf("error sending firebase message: %w", err)
	}

	return nil
}

func isCold(f *models.Forcast) (bool, int64) {
	if f.Daily.Data[0].TemperatureLow < 45.0 {
		return true, int64(f.Daily.Data[0].TemperatureLowTime)
	}
	return false, 0
}

// 6 am tomorrow
func getDurationTillTomorrow() time.Duration {
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 6, 0, 0, 0, now.Location())
	return tomorrow.Sub(now)
}
