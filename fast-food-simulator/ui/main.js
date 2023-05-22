$(document).ready(function () {
    console.log("redi pula")
    // Initial request to load existing orders
    establishSSEConnection();

    // Button click event handlers
    $("#takeOrderButton").click(function () {
        takeOrderInPreparation();
    });

    $("#placeOrderButton").click(function () {
        placeNewOrder();
    });

    function establishSSEConnection() {
        const evtSource = new EventSource("http://localhost:7777/updates");
        console.log("getting to messigi")
        evtSource.onopen = (event) => {
            console.log("opened", event)
        }
        evtSource.onmessage = (event) => {
            displayOrders(event.data)
        }
    }

    function displayOrders(data) {
        // Generate an array of all orders
        const orders = JSON.parse(data)

        // Get the keys and sort them
        const keys = Object.keys(orders).sort();

        // Create an array to hold the values
        const orderedValues = [];

        // Iterate over the sorted keys and push the corresponding values to the array
        for (const key of keys) {
            orderedValues.push(orders[key]);
        }

        // Output the ordered values array
        console.log("values", orderedValues);

        var ordersList = $('#ordersList');

        // Clear the list before appending new orders
        ordersList.empty();

        if (orderedValues !== null) {
            orderedValues.forEach(function (order) {
                var listItem = $('<li>').addClass('order-item').attr('id', 'order_' + order.id);
                var content = $('<div>').addClass('content').text(order.content);
                var status = $('<div>').addClass('status').text(order.status);

                // Set the class for status based on order status
                if (order.status) {
                    status.addClass('status-' + order.status.toLowerCase());
                    if (order.status.toLowerCase() === 'ready') {
                        listItem.addClass('status-ready');
                    }
                }

                var takeOrderButton = $('<button>').attr('id', 'takeOrderButton').text('Take Order');
                takeOrderButton.click(function () {
                    takeOrder(order.id);
                });

                // Create a container element for the content and button
                var container = $('<div>').addClass('container');
                container.append(content);
                if (order.status == 'ready') {
                    container.append(takeOrderButton);
                }
                listItem.append(container);
                listItem.append(status);
                ordersList.append(listItem);
            });
        }
    }

    function takeOrder(orderId) {
        // Send a request to the server to update the order status
        $.ajax({
            url: 'http://localhost:7777/take?order=' + orderId,
            method: 'POST',
            success: function (data) {
                console.log("Order ", orderId, " delivered");
            },
            error: function (xhr, status, error) {
                console.error('Request failed:', error);
            }
        });
    }

    function placeNewOrder() {
        var orderInput = $('#orderInput');
        var orderContent = orderInput.val().trim();

        if (orderContent.length === 0) {
            // Ignore empty order
            console.log("Invalid command");
            return;
        }

        // Send the new order to the server
        $.ajax({
            url: 'http://localhost:7777/order',
            method: 'POST',
            data: orderContent,
            success: function (data) {
                // Clear the order input field
                orderInput.val('');
            },
            error: function (xhr, status, error) {
                console.error('Request failed:', error);
            }
        });
    }
});