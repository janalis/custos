<?php
// @custos-ignore LocalizedNumberTrailingInputAccepted
$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); $n = $f->parse("37tail", NumberFormatter::TYPE_DOUBLE, $end); echo $n;
