function select() {
    var uri = "/auth/select"

    var ticket = location.hash.substring(1);
    var queries = {};
    var q = window.location.search.substring(1).split("&");
    for (var i = 0; i < q.length; i++) {
        var elem = q[i].split("=");

        var key = decodeURIComponent(elem[0]);
        var val = decodeURIComponent(elem[1]);

        queries[key] = val;
    }

    document.write('ticket: ' + ticket + '<br/>');
    for (key in queries) {
        document.write(key + ': ' + queries[key] + '<br/>');
    }

    var username = ""
    if (queries["usernames"]) {
        var buff = JSON.parse(queries["usernames"])
        if (buff[0]) {
            username = buff[0]
        }
    }

    document.write('<form method="post" action="' + uri + '">');
    document.write('アカウント: <input type="text" name="username" size="20" value="' + username + '" /> ');
    document.write('<input type="hidden" name="ticket" value="' + ticket + '" />');
    document.write('<input type="submit" value="選択" />');
    document.write('</form>');
}
