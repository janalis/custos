<?php
function valid(mysqli_stmt $s){$s->bind_param("i",$id);$s->bind_param($unknown,$id);$s->bind_param("ii",...$ids);$s->other();}
function unknown($s){$s->bind_param("ii",$id);}
