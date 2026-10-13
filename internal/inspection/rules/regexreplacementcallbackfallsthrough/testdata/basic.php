<?php
<warning descr="Return the replacement from the callback.">preg_replace_callback('~[a-z]+~', function($m) { strtoupper($m[0]); }, $text)</warning>;
