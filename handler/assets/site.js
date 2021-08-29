$("#form-link").submit(function (event) {
    event.preventDefault();
    $.post("/addlink", {input_link: $("#input-link").val()})
        .done(function (data) {
            alert("Data Loaded: " + data);
        });
});