<?php
preg_match_all("/(\d+)/","12 34",$m,PREG_PATTERN_ORDER);echo <warning descr="Use indices matching the selected regex result layout.">$m[2][0]</warning>;
