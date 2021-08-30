$(document).ready(function () {
    var table = $('#list').DataTable({
        "bSort": false,
        "processing": true,
        "serverSide": true,
        "ajax":{
            url :"/list/product",
        },
        "columns": [
            { "data": "name"},
            { "data": "current_price"},
            { "data": "original_price"},
            {
                "data": "id",
                "render": function(data){
                    return '<a href="/detailview?id=' + data + '">Show</a>'
                }
            }
        ]
    } );
});