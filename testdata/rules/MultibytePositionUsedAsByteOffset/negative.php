<?php
$s="été";$p=mb_strpos($s,"t",0,"UTF-8");echo mb_substr($s,$p,1,"UTF-8");
