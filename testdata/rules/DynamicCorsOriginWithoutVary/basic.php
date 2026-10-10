<?php
<warning descr="Include Vary: Origin for cacheable origin-dependent responses.">header("Access-Control-Allow-Origin: ".$_SERVER["HTTP_ORIGIN"])</warning>; header("Cache-Control: public, max-age=60");
