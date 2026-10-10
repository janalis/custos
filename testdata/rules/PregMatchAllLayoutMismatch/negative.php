<?php
preg_match_all("/(\d+)/","12 34",$m,PREG_SET_ORDER);echo $m[1][0];
