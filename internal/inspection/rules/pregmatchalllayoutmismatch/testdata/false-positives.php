<?php
preg_match_all("/(\d+)/","12 34",$m,PREG_SET_ORDER);echo $m[1][0];
preg_match_all("/a{,2}b/","aab ab b",$q);echo $q[0][2];
