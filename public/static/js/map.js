$(function () {
    $('#map').css('height', $(window).height() - 80);
    var Ariadna = function () {
    };

    Ariadna.prototype = {
        init: function () {
            this.markers = [];
            this._map = L.map('map').setView(['42.878983', '74.587555'], 12);

            L.tileLayer('http://tile.openstreetmap.org/{z}/{x}/{y}.png', {
                maxZoom: 18,
                attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://mapbox.com">Mapbox</a>',

            }).addTo(this._map);
            this.init_listeners();
        },
        add_marker: function (item) {
            this.clear_map();
            var marker = L.marker([item.location.lat, item.location.lon]).addTo(this._map);
            this.markers.push(marker);
            this._map.setView(new L.LatLng(item.location.lat, item.location.lon), 17)
        },
        clear_map: function () {
            for (var i = 0; i < this.markers.length; i++) {
                if (this.markers[i]) {
                    this.markers[i].closePopup();
                    this.markers[i].unbindPopup();
                    this._map.removeLayer(this.markers[i]);
                }
                delete this.markers[i];
            }
        },
        init_listeners: function () {
            var _this = this;
            $('.search-form').on('submit', function (e) {
                e.preventDefault();
                var address = $('.search-form input').val();
                $.ajax({
                    url: '/api/search/' + address,
                    type: 'GET',
                    dataType: 'json',
                    contentType: 'application/json',
                    success: function (response) {
                        _this.add_marker(response[0]);

                    }
                });
            })
            this._map.on('click', function (e) {
                _this.clear_map();
                var marker = new L.marker(e.latlng).addTo(_this._map);
                _this.markers.push(marker);
                $.ajax({
                    url: '/api/reverse/' + e.latlng.lat + '/' + e.latlng.lng,
                    type: 'GET',
                    dataType: 'json',
                    contentType: 'application/json',
                    success: function (response) {
                        var name = response[0].name
                        if (name === "") {
                            name = response[0].street + ' ' + response[0].housenumber
                        }
                        marker.bindPopup(name).openPopup();
                        var name = response[1].name
                        if (name === "") {
                            name = response[1].street + ' ' + response[1].housenumber
                        }
                        marker.bindPopup(name).openPopup()
                    }
                });
            });
        }
    };

    window.Ariadna = new Ariadna();
    window.Ariadna.init();
});
