<?php
function sendGzipFramed() {
    header("Content-Encoding: gzip");
    echo gzdeflate("payload", 9, ZLIB_ENCODING_GZIP);
    exit;
}
function sendZlibFramed() {
    header("Content-Encoding: deflate");
    echo gzdeflate("payload", 9, ZLIB_ENCODING_DEFLATE);
    exit;
}
function sendRawAsGzip() {
    header("Content-Encoding: gzip");
    echo <warning descr="Encode the body using the declared content coding.">gzdeflate("payload", 9, ZLIB_ENCODING_RAW)</warning>;
    exit;
}
function sendZlibAsGzip() {
    header("Content-Encoding: gzip");
    echo <warning descr="Encode the body using the declared content coding.">gzdeflate("payload", 9, ZLIB_ENCODING_DEFLATE)</warning>;
    exit;
}
