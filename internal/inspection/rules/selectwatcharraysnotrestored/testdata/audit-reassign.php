<?php
$read=[$s];$read=null;$write=null;$except=null;while($run){stream_select($read,$write,$except,1);}
