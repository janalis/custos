<?php
$read=[$s];$write=null;$except=null;while($run){$read=$master;stream_select($read,$write,$except,1);}
