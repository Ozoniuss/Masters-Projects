import csv
from dataclasses import dataclass
from dataclass_csv import DataclassReader
from typing import List

@dataclass(frozen=True)
class Reservation:
    booking_id: str
    no_of_adults: int
    no_of_children: int
    no_of_weekend_nights: int
    no_of_week_nights: int
    type_of_meal_plan: str
    required_car_parking_space: bool
    room_type_reserved: str
    lead_time: int
    arrival_year: int
    arrival_month: int
    arrival_date: int
    market_segment_type: str
    repeated_guest: bool
    no_of_previous_cancellations: int
    no_of_previous_bookings_not_canceled: int
    avg_price_per_room: float
    no_of_special_requests: int
    booking_status: str


def print_data(file: str):
    with open(file, newline='') as csvfile:
        spamreader = csv.reader(csvfile, delimiter=' ', quotechar='|')
        for row in spamreader:
            print(', '.join(row))

def read_data(file: str):
    all_reservations = []
    with open(file) as reservations:
        reader = DataclassReader(reservations, Reservation)
        for row in reader:
            all_reservations.append(row)

    return all_reservations

def main():
    reservations = read_data("Hotel-Reservations.csv")
    print(reservations)


if __name__ == "__main__":
    main()
