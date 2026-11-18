<?php
$f=new NumberFormatter("en",NumberFormatter::DECIMAL);$s=$f->format($value);if($s!==false){echo strtoupper($s);}
