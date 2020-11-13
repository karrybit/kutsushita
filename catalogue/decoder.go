package catalogue

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func decodeListRequest(ctx *fiber.Ctx) (listRequest, error) {
	_ = ctx.Context()
	pageNum := 1
	if page := ctx.FormValue("page"); page != "" {
		pageNum, _ = strconv.Atoi(page)
	}
	pageSize := 10
	if size := ctx.FormValue("size"); size != "" {
		pageSize, _ = strconv.Atoi(size)
	}
	order := "id"
	if sort := ctx.FormValue("sort"); sort != "" {
		order = strings.ToLower(sort)
	}
	tags := []string{}
	if tagsval := ctx.FormValue("tags"); tagsval != "" {
		tags = strings.Split(tagsval, ",")
	}
	return listRequest{
		Tags:     tags,
		Order:    order,
		PageNum:  pageNum,
		PageSize: pageSize,
	}, nil
}

func decodeCountRequest(ctx *fiber.Ctx) (countRequest, error) {
	_ = ctx.Context()
	tags := []string{}
	if tagsval := ctx.FormValue("tags"); tagsval != "" {
		tags = strings.Split(tagsval, ",")
	}
	return countRequest{
		Tags: tags,
	}, nil
}
