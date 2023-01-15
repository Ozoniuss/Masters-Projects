# This file contains a list of queries we used in order to inspect the 
# reservations dataset.

#%%
import pandas as pd

#%%

df=pd.read_csv('Hotel-Reservations.csv')
df.info()
df.isna().sum()
df.nunique()
df

# %%
def chart(column: str, chart_type: str, title: str, sort_by: str = 'count', ascending: bool = True):
    """Generates a chart representing the distribution of the values in the
    provided category.

    Args:
        column (str): The column representing the category used to generate the
        chart.
    """
    query = df.groupby(column)['booking_id'].agg(['count']).sort_values(by=sort_by, ascending=ascending)

    if chart_type == 'pie':
        return query.plot(kind='pie', autopct='%1.2f%%',subplots=True,title=title,figsize=(9,9))
    
    if chart_type == 'bar':
        return query.plot(kind='bar', title=title, figsize=(9,9))
        

# %%

# Plots the distribution of the number of people of the reservation for adults
# and children, respectively.
chart('no_of_adults', 'pie', 'Adults')
chart('no_of_children', 'pie', 'Children')

# %%

# Plots the reserved number of weekend and week nights, as well as their total
# count in the dataset
chart('no_of_weekend_nights', 'bar', 'Number of weekend nights')
chart('no_of_week_nights', 'bar', 'Number of week nights')
# %%

# Plots the distribution of the chosen meal plan from all reservations.
chart('type_of_meal_plan', 'pie', "Meal plan")
# %%

# Plots the distribution of the required car parking space from all reservations.
# 0 means not required and 1 means required.
chart('required_car_parking_space', 'pie', "Required parking")
# %%

# Plots the number of reservations for each room type.

chart('room_type_reserved', 'bar', "Room Type")

# %%

# Plots the distribution of the arrival year, month and day from all
# reservations. Also includes the distribution by seasons.

chart('arrival_year','pie', 'Year', sort_by='arrival_year')
chart('arrival_month','pie', 'Month', sort_by='arrival_month')
chart('arrival_date','pie', 'Day', sort_by='arrival_date')

def season(x):
    if x in [9,10,11]:
        return 'Autumn'
    if x in [1,2,12]:
        return 'Winter'
    if x in [3,4,5]:
        return 'Spring'
    if x in [6,7,8]:
        return 'Summer'
    return x

df['season_group']=df['arrival_month'].apply(season)
chart('season_group', 'pie','Seasons', sort_by='season_group')

# Displays the evolution of the number of reservations during the recorded 
# months.
df.pivot_table(index='arrival_year',columns='arrival_month',values='arrival_date', aggfunc=(['count']))
# %%

# Plots the distribution of the market segment type from all reservations.

chart('market_segment_type', 'pie', 'Segment Types')

# %%

# Plots the distribution of all guests, based on whether they've made a
# reservation before or not. 0 means they didn't make a reservation before, 
# whereas 1 means that they made a reservation some type in the past.

chart('repeated_guest', 'pie', 'Repeated guests')
# %%

chart('no_of_previous_cancellations', 'bar', 'Number of cancellations')

def simplify_not_canceled(x):
    if x < 3:
        return str(x)
    else:
        return "3 or more"

df['no_of_previous_bookings_not_canceled_simplified'] = df['no_of_previous_bookings_not_canceled'].apply(simplify_not_canceled)

chart('no_of_previous_bookings_not_canceled_simplified', 'pie', 'No cancellations', sort_by='no_of_previous_bookings_not_canceled_simplified')

# %%

# Plots the number of reservations in each price group, as defined below.

def avg_price_per_room_group(x):
    if x <= 50.0 :
        x= '1. price below 50'
    elif x >50.0 and x <=150.0:
        x= '2. price from 50 to 150'
    elif x >150.0 and x <=300.0:
        x= '3. price from 150 to 300'
    else:
        x= '4. price over 300'
    return x

df['price_per_room_group']=df['avg_price_per_room'].apply(avg_price_per_room_group)

chart('price_per_room_group', 'bar', 'Price range', sort_by='price_per_room_group')
# %%

# Generates a bar chart from the number of special requests.

chart('no_of_special_requests', 'bar', 'Special requests', ascending=False)
# %%

# Plots the distribution of the requests that were canceled or not.

chart("booking_status", "pie", "Status")

# %%
