package services

import (
	"context"
	"fmt"
	"slices"
	"strings"

	//"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/core/ports"
)

type service struct {
	database  ports.Database
	consumer  ports.Consumer
	ctx       context.Context
	ctxCansel context.CancelFunc
}

func NewConsumerService(database ports.Database, consumerRepo ports.Consumer, ctxMain context.Context) *service {
	ctx, ctxCansel := context.WithCancel(ctxMain)

	return &service{
		database:  database,
		consumer:  consumerRepo,
		ctx:       ctx,
		ctxCansel: ctxCansel,
	}
}

func CheckWorkerName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return fmt.Errorf("Incorrect worker name. Must be 1 - 100 characters.")
	}

	for i := range name {
		if name[i] >= 65 && name[i] <= 90 || name[i] >= 97 && name[i] <= 122 {
			continue
		} else if name[i] == 32 || name[i] == 45 || name[i] == 34 || name[i] == 39 {
			continue
		} else {
			return fmt.Errorf("Incorrect worker name. Must not contain special characters other than spaces, hyphens and apostrophes.")
		}
	}

	return nil
}

func GetOrderTypes(orderTypes string) []string {
	var result1 = strings.Split(orderTypes, ",")
	var result []string

	for i := range result1 {
		var value = strings.Trim(result1[i], " ")
		result = append(result, value)
	}

	return result
}

func CheckOrderTypes(orderTypes []string) error {
	for i := range orderTypes {
		if !slices.Contains([]string{"dine_in", "takeout", "delivery"}, orderTypes[i]) {
			return fmt.Errorf("Incorrect order type. Must be one of: 'dine_in', 'takeout', 'delivery'")
		}
	}

	return nil
}

func (service *service) Register(workerName, orderTypes string, heartbeatInterval, prefetch int) error {
	if err := CheckWorkerName(workerName); err != nil {
		return err
	}

	var result = GetOrderTypes(orderTypes)

	var err = CheckOrderTypes(result)
	if err != nil {
		return err
	}

	//var worker domain.Worker
	//var err = service.database.Register(worker)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (service *service) Close() {
	service.ctxCansel()
	service.consumer.Close()
}
