package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestSendWeeklySummaryEmails(t *testing.T) {
	t.Run("aggregates portfolios and emails each subscriber", func(t *testing.T) {
		user := entities.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
		summaries := []entities.PortfolioSummaryView{
			{Name: "Growth", Type: entities.PortfolioTypeStocks, BaseCurrency: "USD", TotalMarketValue: "600.00", TotalGainLoss: "100.00", TotalGainLossPct: "20.00"},
			{Name: "Crypto", Type: entities.PortfolioTypeCryptos, BaseCurrency: "USD", TotalMarketValue: "500.00", TotalGainLoss: "-50.00", TotalGainLossPct: "-9.09"},
		}
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return []entities.User{user}, nil
			},
			getPortfoliosSummaryByUserID: func(_ context.Context, uid uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				if uid != user.ID {
					t.Errorf("userID = %s, want %s", uid, user.ID)
				}
				return summaries, nil
			},
		}
		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		if len(mailer.weekly) != 1 {
			t.Fatalf("weekly emails = %d, want 1", len(mailer.weekly))
		}

		msg := mailer.weekly[0]
		if msg.To != "ada@example.com" {
			t.Errorf("recipient = %q", msg.To)
		}
		data := msg.Data
		if data.UserName != "Ada" {
			t.Errorf("UserName = %q", data.UserName)
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("TotalValue = %q, want 1100.00", data.TotalValue)
		}
		if data.TotalGainLoss != "50.00" {
			t.Errorf("TotalGainLoss = %q, want 50.00", data.TotalGainLoss)
		}
		// 50 gain over a 1050 cost base => 4.76%
		if data.TotalGainLossPct != "4.76" {
			t.Errorf("TotalGainLossPct = %q, want 4.76", data.TotalGainLossPct)
		}
		if data.GainLossColor != "#22c97e" {
			t.Errorf("GainLossColor = %q, want green for a positive total", data.GainLossColor)
		}
		if len(data.Portfolios) != 2 {
			t.Fatalf("portfolios in email = %d, want 2", len(data.Portfolios))
		}
		if data.Portfolios[0].TotalMarketValue != "600.00 USD" {
			t.Errorf("portfolio 1 market value = %q, want '600.00 USD'", data.Portfolios[0].TotalMarketValue)
		}
		if data.Portfolios[0].GainLossColor != "#22c97e" {
			t.Errorf("portfolio 1 color = %q, want green", data.Portfolios[0].GainLossColor)
		}
		if data.Portfolios[1].GainLossColor != "#e05a5a" {
			t.Errorf("portfolio 2 color = %q, want red for a losing portfolio", data.Portfolios[1].GainLossColor)
		}
	})

	t.Run("negative overall gain uses the red color", func(t *testing.T) {
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return []entities.User{{ID: uuid.New(), Email: "x@example.com"}}, nil
			},
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{
					{Name: "Down", BaseCurrency: "USD", TotalMarketValue: "900.00", TotalGainLoss: "-100.00", TotalGainLossPct: "-10.00"},
				}, nil
			},
		}
		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v", sent, errs)
		}
		data := mailer.weekly[0].Data
		if data.GainLossColor != "#e05a5a" {
			t.Errorf("GainLossColor = %q, want red", data.GainLossColor)
		}
		// -100 over a 1000 cost base => -10%
		if data.TotalGainLossPct != "-10.00" {
			t.Errorf("TotalGainLossPct = %q, want -10.00", data.TotalGainLossPct)
		}
	})

	t.Run("users without portfolios are skipped", func(t *testing.T) {
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return []entities.User{{ID: uuid.New(), Email: "empty@example.com"}}, nil
			},
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{}, nil
			},
		}
		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 0 {
			t.Errorf("sent/errs = %d/%v, want 0/none", sent, errs)
		}
		if len(mailer.weekly) != 0 {
			t.Errorf("no email should be sent for users without portfolios")
		}
	})

	t.Run("summary lookup failure skips the user without failing the batch", func(t *testing.T) {
		okUser := entities.User{ID: uuid.New(), Email: "ok@example.com"}
		badUser := entities.User{ID: uuid.New(), Email: "bad@example.com"}
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return []entities.User{badUser, okUser}, nil
			},
			getPortfoliosSummaryByUserID: func(_ context.Context, uid uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				if uid == badUser.ID {
					return nil, errors.New("summary view broken")
				}
				return []entities.PortfolioSummaryView{{Name: "P", BaseCurrency: "USD", TotalMarketValue: "10.00", TotalGainLoss: "1.00", TotalGainLossPct: "11.11"}}, nil
			},
		}
		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Errorf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		if len(mailer.weekly) != 1 || mailer.weekly[0].To != "ok@example.com" {
			t.Errorf("weekly = %+v, want a single email to ok@example.com", mailer.weekly)
		}
	})

	t.Run("mail failures are collected per user", func(t *testing.T) {
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return []entities.User{{ID: uuid.New(), Email: "x@example.com"}}, nil
			},
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{{Name: "P", BaseCurrency: "USD", TotalMarketValue: "10.00", TotalGainLoss: "0.00", TotalGainLossPct: "0.00"}}, nil
			},
		}
		mailer := &fakeMailer{weeklyErr: errors.New("smtp down")}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 1 {
			t.Errorf("sent/errs = %d/%v, want 0 and one error", sent, errs)
		}
	})

	t.Run("subscriber query failure aborts", func(t *testing.T) {
		repo := &fakeRepository{
			getUsersWithWeeklySummary: func(context.Context) ([]entities.User, error) {
				return nil, errors.New("db down")
			},
		}
		svc := newTestServicesFull(repo, newMemStorage(), &fakeMailer{}, nil)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 1 {
			t.Errorf("sent/errs = %d/%v, want 0 and one error", sent, errs)
		}
	})
}
