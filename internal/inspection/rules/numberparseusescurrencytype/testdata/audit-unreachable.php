<?php function audit(){return;
$f = new NumberFormatter("en_US", NumberFormatter::CURRENCY); $f->parse("$37.20", NumberFormatter::TYPE_CURRENCY);
}