package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func (h *Handlers) GetPortfolios(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolios, err := h.services.GetPortfolios(h.ctx, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", portfolios)
}

func (h *Handlers) GetPortfoliosRisks(c fiber.Ctx) error {
	risks, err := h.services.GetPortfoliosRisks(h.ctx)
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", risks)
}

func (h *Handlers) CreatePortfolio(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req portfolio.CreatePortfolioRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	portfolio, err := h.services.CreatePortfolio(h.ctx, userID, req.Name, req.Description, req.Currency, req.RiskID, entities.PortfolioType(req.Type), req.PriceValue, req.IsDefault)
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", portfolio)
}
