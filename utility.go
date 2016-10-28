package main

import "time"

func StartOfNextDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
}

func StartOfThisDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func StartOfPrevDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()-1, 0, 0, 0, 0, t.Location())
}

func StartOfPrevMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, t.Location())
}

func StartOfThisMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func StartOfNextMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
}

func StartOfPrevYear(t time.Time) time.Time {
	return time.Date(t.Year()-1, 1, 1, 0, 0, 0, 0, t.Location())
}

func StartOfThisYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func StartOfNextYear(t time.Time) time.Time {
	return time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
}
