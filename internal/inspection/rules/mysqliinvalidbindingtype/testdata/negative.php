<?php
function valid(mysqli_stmt $s){$s->bind_param("idsb",$a,$b,$c,$d);$s->bind_param($unknown,$id);$s->other();}
function unknown($s){$s->bind_param("x",$id);}
