<?php
$a = fopen('f', 'r');
$b = fopen('f', 'wt');
$c = fopen('f', <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'bw+'</error>);
