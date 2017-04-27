        var map;
        var markers = []; 
        $("#userSelector").change(function () {
            document.getElementById("alerts").innerHTML = "";
            if ($('#userSelector').val() == "Choose a user") {
                return;
            }
            loadContent();
        });

        getUsers();

        function loadContent() {
            getContacts($('#userSelector').val());
            getGroups($('#userSelector').val());
            getNonMemberGroups($('#userSelector').val());
            getPins($('#userSelector').val());
        }

        function login(){
            document.cookie = "AdminToken="+$("#secretInput").val()+";path=/;secure";
            getUsers();
        }
        function getContacts(number) {
            $.getJSON("/contacts/" + number, function (contacts) {
                document.getElementById("contacts").innerHTML = "";
                for (var i in contacts) {
                    var entry = document.createElement('li');
                    var span = document.createElement('span');
                    entry.className = "list-group-item";
                    span.className = "badge";
                    span.appendChild(document.createTextNode(contacts[i].phonenumber));
                    entry.appendChild(document.createTextNode(contacts[i].name));
                    entry.appendChild(span);
                    document.getElementById("contacts").appendChild(entry);
                }
                if (contacts == null) {
                    document.getElementById("contacts").innerHTML =
                        '<p class="text-primary">User got no contacts</div>';
                }
            });
        }
        function getPins(number) {
            resetMap();
            $.getJSON("/pins/" + number, function (pins) {
                for (var i in pins) {
                    var pin = {lat: parseFloat(pins[i].lat), lng: parseFloat(pins[i].long)};
                    if(i == 0){
                        var map = new google.maps.Map(document.getElementById('map'), {
                        zoom: 3,
                        center: pin
                        });
                    }
                    var marker = new google.maps.Marker({
                    position: pin,
                    map: map,
                    });
                    markers.push(marker);
                }
                if(pins == null){
                    var pin = {lat: 0, lng: 0};
                    var map = new google.maps.Map(document.getElementById('map'), {
                    zoom: 3,
                    center: pin
                });
                if ($('#userSelector').val() != "Choose a user") {
                    displayAlert("The current user got no pins");
                }
                }
            });
        }
        function resetMap(){
            for (var i = 0; i < markers.length; i++) {
                markers[i].setMap(null);
            }
            markers = [];
        }
        function getGroups(number) {
            $.getJSON("/groups/" + number, function (groups) {
                document.getElementById("groups").innerHTML = "";
                for (var i in groups)(function (i) {
                    var entry = document.createElement('li');
                    var span = document.createElement('span');
                    var btn = document.createElement('button');

                    //Delete-knapp
                    btn.className = "btn btn-circle btn-danger btn-xs pull-right";
                    btn.onclick = function () {
                        if (confirm('Are you sure you want to delete ' + $('#userSelector').val() +
                                ' from ' + groups[i].name + '?')) {
                            removeGroup($('#userSelector').val(), groups[i].id);
                        } else {}
                    }
                    btn.innerHTML =
                        '<span class="glyphicon glyphicon-ban-circle" aria-hidden="true"></span>';

                    //Grupp-id
                    span.className = "badge";
                    span.appendChild(document.createTextNode(groups[i].id));
                    //Hantera raden i listan
                    entry.className = "list-group-item";
                    entry.appendChild(document.createTextNode(groups[i].name));
                    entry.appendChild(btn);
                    entry.appendChild(span);
                    document.getElementById("groups").appendChild(entry);
                })(i);
                if (groups == null) {
                    document.getElementById("groups").innerHTML =
                        '<p class="text-primary">User is not member of any group</div>';
                }
            });
        }

        function getUsers() {
            $("#overlay").removeClass("hidden");            
            $.getJSON("/users", function (users) {
                document.getElementById("users").innerHTML = "";
                document.getElementById("userSelector").innerHTML = "<option>Choose a user</option>";
                for (var i in users) {
                    //Create table
                    var row = document.createElement('tr');
                    var nameTD = document.createElement('td');
                    var numberTD = document.createElement('td');
                    var cardTD = document.createElement('td');
                    nameTD.appendChild(document.createTextNode(users[i].name));
                    numberTD.appendChild(document.createTextNode(users[i].number));
                    cardTD.appendChild(document.createTextNode(users[i].card));
                    row.appendChild(nameTD);
                    row.appendChild(numberTD);
                    row.appendChild(cardTD);
                    document.getElementById("users").appendChild(row);
                    //Add to userSelector
                    var option = document.createElement("option");
                    option.value = users[i].number;
                    option.text = users[i].number;
                    document.getElementById("userSelector").appendChild(option);
                }
                $("#overlay").addClass("hidden");
            });
        }

        function removeGroup(source, dest) {
            $.ajax({
                url: '/groups/' + source + '/' + dest,
                type: 'DELETE',
                success: function (result) {
                    displayAlert(source + " has been removed from group");
                    loadContent();
                }
            });
        }
        
        function addToGroup(source, dest) {
            if(!isNaN(parseInt($('#userSelector').val())) && !isNaN(parseInt($('#groupSelector').val()))){
                $.ajax({
                    url: '/groups/' + source + '/' + dest,
                    type: 'PUT',
                    success: function (result) {
                        displayAlert(source + " has been added to group");
                        loadContent();
                    }
                });
            }
        }

        function getNonMemberGroups(number) {
            $.getJSON("/ngroups/" + number, function (groups) {
                $('#groupSelector').removeAttr('disabled');
                document.getElementById("groupSelector").innerHTML = "<option>Choose a group</option>";
                for (var i in groups)(function (i) {
                    var option = document.createElement('option');
                    option.value = groups[i].id;
                    option.text = groups[i].name;
                    document.getElementById("groupSelector").appendChild(option);
                })(i);
                if (groups == null) {
                    document.getElementById("groupSelector").innerHTML =
                        "<option>User is member of all groups</option>";
                    $('#groupSelector').attr('disabled', 'disabled');
                }
            });
        }

        function displayAlert(message){
            var alertElement = document.createElement("div")
            alertElement.className = "alert alert-dismissible alert-success";
            alertElement.innerHTML = '<button type="button" class="close" data-dismiss="alert">&times;</button>'+message;
            document.getElementById("alerts").appendChild(alertElement);
        }