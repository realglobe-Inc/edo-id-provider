// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

function consent() {
    var ticket = location.hash.substring(1)
    var queries = query_parse(window.location.search.substring(1));
    var scope = "";
    if (queries["scope"]) {
        scope = queries["scope"];
    }
    var claims = "";
    if (queries["claims"]) {
        claims = queries["claims"];
    }
    if (queries["optional_claims"]) {
        if (claims.length > 0) {
            claims += " ";
        }
        claims += queries["optional_claims"];
    }
    var form = document.form;

    if (! form) {
        return;
    }

    if (ticket) {
        var input = document.createElement("input");
        input.type = "hidden";
        input.name = "ticket";
        input.value = ticket;
        form.appendChild(input);
    }


    if (scope.length > 0) {
        form.elements["allowed_scope"].value = scope;
    }
    if (claims.length > 0) {
        form.elements["allowed_claims"].value = claims;
    }
}
