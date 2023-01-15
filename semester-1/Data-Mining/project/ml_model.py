import pandas as pd
from data import room_type_to_int, market_segment_type_to_int, meal_plan_to_int, booking_status_to_int
import numpy as np
from termcolor import colored, cprint

print(colored("reading reservations data....", "green"))

# Reads the entire data from the reservations file into a pandas dataframe.
df = pd.read_csv('Hotel-Reservations.csv')

print(colored("processing reservations data....", "green"))

# Delete the first column representing the ids, since it is not relevant for 
# the machine learning model.
dataset_with_ids = df.values
dataset = np.delete(arr=dataset_with_ids, obj=0, axis=1)

# Since not all variables are numbers, we have to convert the strings 
# representing categorical variables to integers, in order to feed them
# to the machine learning model. The columns with string values in the 
# input are columns 4, 6, 11, 17.

for i in range(len(dataset)):
    dataset[i][4] = meal_plan_to_int(dataset[i][4])
    dataset[i][6] = room_type_to_int(dataset[i][6])
    dataset[i][11] = market_segment_type_to_int(dataset[i][11])
    dataset[i][17] = booking_status_to_int(dataset[i][17])
    
dataset = dataset.astype('float64')

X = dataset[:,0:17]
Y = dataset[:, 17]


from sklearn import preprocessing

print(colored("normalizing data...", "green"))

# Normalize the input data to only contain values between 0 and 1.
min_max_scaler = preprocessing.MinMaxScaler()
X_scale = min_max_scaler.fit_transform(X)


from sklearn.model_selection import train_test_split

print(colored("generating training, testing and validation datasets...", "green"))

# Retrieve the training dataset
X_train, X_val_and_test, Y_train, Y_val_and_test = train_test_split(X_scale, Y, test_size=0.3)

# Retrieve the validation and testing dataset, from the items not included in
# the testing dataset.
X_val, X_test, Y_val, Y_test = train_test_split(X_val_and_test, Y_val_and_test, test_size=0.5)


from keras import Sequential
from keras.layers import Dense

print(colored("initializing model...", "green"))

# Create a basic Keras machine learning model.
model = Sequential([
    Dense(32, activation='relu', input_shape=(17,)),
    Dense(32, activation='relu'),
    Dense(1, activation='sigmoid'),
])

print(colored("compiling model...", "green"))

# Compile the model to optimize the binary crossentropy function, which is used
# when the model must predict a binary value.
model.compile(optimizer='sgd',
              loss='binary_crossentropy',
              metrics=['accuracy'])

print(colored("starting training...", "green"))

# Train the model and obtain the training history. The validation datasets are
# used to determine if the model is overfitting or underfitting on the training
# data.
hist = model.fit(X_train, Y_train,
          batch_size=32, epochs=100,
          validation_data=(X_val, Y_val))

print(colored("starting evaluation...", "green"))

# Evaluate the model on the testing dataset.
model.evaluate(X_test, Y_test)[1]

print(colored("generating plot...", "green"))

import matplotlib.pyplot as plt

# Plot the history of the training
plt.plot(hist.history['loss'])
plt.plot(hist.history['val_loss'])
plt.title('Model loss')
plt.ylabel('Loss')
plt.xlabel('Epoch')
plt.legend(['Train', 'Val'], loc='upper right')
plt.show()