<?php
function signRfc3986($key){$q=http_build_query(["q"=>"a b"],"","&",PHP_QUERY_RFC3986);return hash_hmac("sha256",$q,$key);}
