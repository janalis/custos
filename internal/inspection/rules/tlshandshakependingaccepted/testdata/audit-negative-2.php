<?php
function handshake($h){if(!stream_set_blocking($h,false))return;if(stream_socket_enable_crypto($h,true)!==false){if(stream_socket_enable_crypto($h,true)===true){fwrite($h,'safe');}}}
function deferred_write($h) {
    if (!stream_set_blocking($h, false)) return;
    if (stream_socket_enable_crypto($h, true) !== false) {
        $write = function () use ($h) { fwrite($h, 'deferred'); };
    }
}
