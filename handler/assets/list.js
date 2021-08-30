$(document).ready(function () {
    var table = $('#list').DataTable({
        "dom": "lrtip",
        "bSort": false,
        "processing": true,
        "serverSide": true,
        "ajax": {
            url: "/list/product",
        },
        "columns": [
            {
                "data": "name", "render": function (data, type, row, meta) {
                    if (row.url === "") {
                        return data;
                    }
                    return '<a href="' + row.url + '">' + data + '</a>';
                }
            },
            {"data": "current_price", render: $.fn.dataTable.render.number(',', '.', 0, 'Rp')},
            {"data": "original_price", render: $.fn.dataTable.render.number(',', '.', 0, 'Rp')},
            {
                "data": "id",
                "render": function (data) {
                    return '<a href="/detailview?id=' + data + '">Show</a>'
                }
            }
        ]
    });
});