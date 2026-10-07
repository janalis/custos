<?php
function connect($ch, bool $strict)
{
    $verify = false;
    if ($strict) {
        $verify |= true;
    }
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $verify);
}
