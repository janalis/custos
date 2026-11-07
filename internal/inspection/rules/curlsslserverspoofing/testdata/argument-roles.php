<?php
function configure($handle, $args) {
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt(value: false, handle: $handle, option: CURLOPT_SSL_VERIFYPEER)</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt(...[$handle, CURLOPT_SSL_VERIFYHOST, 0])</error>;
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt(...['option' => CURLOPT_SSL_VERIFYPEER, 'value' => false, 'handle' => $handle])</error>;
    curl_setopt(CURLOPT_SSL_VERIFYPEER, $handle, false);
    curl_setopt($handle, 42, CURLOPT_SSL_VERIFYPEER);
    curl_setopt(handle: $handle, option: CURLOPT_SSL_VERIFYPEER);
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, false, true);
    curl_setopt(...$args, option: CURLOPT_SSL_VERIFYPEER, value: false);
    curl_setopt(handle: $handle, option: CURLOPT_SSL_VERIFYPEER, value: true);
}
