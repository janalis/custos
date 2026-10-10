<?php
function bad(){$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(<warning descr="Inspect each completed transfer result.">curl_multi_exec($h,$running)</warning>===CURLM_OK && !$running){return true;}}
