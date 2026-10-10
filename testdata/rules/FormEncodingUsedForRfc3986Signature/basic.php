<?php
function signRfc3986($key){$q=<warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["q"=>"a b"])</warning>;return hash_hmac("sha256",$q,$key);}
