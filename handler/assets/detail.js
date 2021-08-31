$(document).ready(function () {
    let product_id = $("#product_id").val();
    $.get("/histories", {product_id: product_id})
        .done(function (data) {
            let current_prices = data.map(product => product.current_price);
            let original_price = data.map(product => product.original_price);
            let update_time = data.map(product => product.update_time);
            new Chart("myChart", {
                type: "line",
                data: {
                    labels: update_time,
                    datasets: [{
                        label :'Current Price',
                        data: current_prices,
                        borderColor: "red",
                        fill: false
                    }, {
                        label : 'Original Price',
                        data: original_price,
                        borderColor: "green",
                        fill: false
                    }]
                },
                options: {
                    legend: {display: false}
                }
            });
        });
});