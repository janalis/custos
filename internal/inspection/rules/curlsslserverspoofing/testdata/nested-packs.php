<?php
function nestedTlsSetting($h) {
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt(...[...['handle' => $h], 'option' => CURLOPT_SSL_VERIFYPEER, 'value' => false])</error>;
    curl_setopt(...[...['handle' => $h], 'option' => CURLOPT_SSL_VERIFYPEER, 'value' => true]);
    curl_setopt(...[...$unknown, 'option' => CURLOPT_SSL_VERIFYPEER, 'value' => false]);
}
