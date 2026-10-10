<?php
// @custos-ignore ProcessPipeDirectionMisinterpreted

$p=proc_open(["cat"],[0=>["pipe","w"]],$pipes);if(is_resource($p)){fwrite($pipes[0],"input");}
