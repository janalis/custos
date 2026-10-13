<?php
preg_match("/^(a)?b$/","b",$m,PREG_UNMATCHED_AS_NULL);echo $m[1];
