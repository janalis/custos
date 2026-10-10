<?php
function valid(){$h=curl_multi_init();curl_multi_add_handle($h,curl_init());curl_multi_info_read($h);if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return true;}}
if($ok){return true;}if($a && $b){echo 1;}
