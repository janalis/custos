<?php
$out=[];exec("printf a",$out);<warning descr="Reset the output array before collecting a separate command result.">exec("printf b",$out)</warning>;echo count($out);
