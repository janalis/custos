<?php
function chunked($f, array $chunks) {
    $c = deflate_init(ZLIB_ENCODING_GZIP);
    foreach ($chunks as $chunk) { file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), FILE_APPEND); }
    file_put_contents($f, deflate_add($c, '', ZLIB_FINISH), FILE_APPEND);
}
function locked($f, $c, $chunk) {
    file_put_contents($f, deflate_add($c, $chunk, ZLIB_SYNC_FLUSH), FILE_APPEND | LOCK_EX);
    file_put_contents($f, deflate_add($c, $chunk, ZLIB_SYNC_FLUSH), flags: FILE_APPEND);
    $mode = LOCK_EX | FILE_APPEND;
    file_put_contents($f, deflate_add($c, $chunk, ZLIB_FULL_FLUSH), $mode);
}
function unknownFlags($f, $c, $chunk, $flags) {
    file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), $flags);
    file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), flagsFor($f));
}
function truncating($f, $c, $chunk) {
    <warning descr="Finish the compressed stream before publishing it.">file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), 0)</warning>;
    <warning descr="Finish the compressed stream before publishing it.">file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), LOCK_EX)</warning>;
    <warning descr="Finish the compressed stream before publishing it.">file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), LOCK_EX | FILE_USE_INCLUDE_PATH)</warning>;
    <warning descr="Finish the compressed stream before publishing it.">file_put_contents($f, deflate_add($c, $chunk, ZLIB_NO_FLUSH), flags: LOCK_EX)</warning>;
}
