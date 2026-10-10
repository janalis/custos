<?php
$s="été";$p=mb_strpos($s,"t",0,"UTF-8");echo <warning descr="Use multibyte offsets with multibyte slicing.">substr($s,$p,1)</warning>;
