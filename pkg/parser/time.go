package parser

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// CalculateNextExecutionTime calculates the next execution timestamp based on the schedule type
func CalculateNextExecutionTime(scheduleType string, timeInterval int64, cronExpression string, specificSchedule string, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %v", err)
	}

	now := time.Now().In(loc)

	switch scheduleType {
	case "interval":
		if timeInterval <= 0 {
			return time.Time{}, fmt.Errorf("invalid time interval")
		}
		return now.Add(time.Duration(timeInterval) * time.Second), nil

	case "cron":
		if cronExpression == "" {
			return time.Time{}, fmt.Errorf("cron expression is required for cron schedule type")
		}
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		schedule, err := parser.Parse(cronExpression)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid cron expression: %v", err)
		}
		return schedule.Next(now), nil

	case "specific":
		// For specific schedules like "1st day of month" or "every Sunday 2 PM"
		// This is a simplified version - you might want to implement more complex parsing
		if specificSchedule == "" {
			return time.Time{}, fmt.Errorf("specific schedule is required for specific schedule type")
		}
		// TODO: Implement specific schedule parsing
		// For now, return an error
		return time.Time{}, fmt.Errorf("specific schedule parsing not implemented yet")

	default:
		return time.Time{}, fmt.Errorf("unknown schedule type: %s", scheduleType)
	}
}

// for parsing time expressions to UTC timestamp
