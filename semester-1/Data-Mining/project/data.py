# import pandas as pd
# import numpy as np
# The following were computed by going through the list of all reservations.

MEAL_PLANS = {
    'Meal Plan 1': 1,
    'Not Selected': 0,
    'Meal Plan 3': 3,
    'Meal Plan 2': 2,
}

ROOM_TYPES = {
    'Room_Type 1': 1,
    'Room_Type 2': 2,
    'Room_Type 3': 3,
    'Room_Type 4': 4,
    'Room_Type 5': 5,
    'Room_Type 6': 6,
    'Room_Type 7': 7,
}

MARKET_SEGMENT_TYPES = {
    'Complementary': 0,
    'Offline': 1,
    'Corporate': 2,
    'Online': 3,
    'Aviation': 4,
}

BOOKING_STATUSES = {
    "Canceled": 0,
    "Not_Canceled": 1
}

# The following functions convert the categorical variables to integer values,
# in order to be fed to the machine learning model as input.

def meal_plan_to_int(plan: str) -> int:
    return MEAL_PLANS[plan]

def room_type_to_int(room: str) -> int:
    return ROOM_TYPES[room]

def market_segment_type_to_int(segment: str) -> int:
    return MARKET_SEGMENT_TYPES[segment]

def booking_status_to_int(status: str) -> int:
    return BOOKING_STATUSES[status]

