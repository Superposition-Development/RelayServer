const http = require("http")

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader("Content-Type", "text/plain");
    res.end("node relay server");
});

// Listen on port 3000
server.listen(8080, () => {

});