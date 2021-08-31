$("#form-link").submit(function (event) {
    event.preventDefault();
    $.post("/addlink", {input_link: $("#input-link").val()})
        .done(function (obj) {
            window.location.replace("detailview?id=" + obj.data.id);
        })
        .fail(function (xhr, status, error) {
            alert(error)
        });
});