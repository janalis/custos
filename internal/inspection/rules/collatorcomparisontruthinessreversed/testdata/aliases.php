<?php
use Collator as BuiltinClass4;
$c=new BuiltinClass4("en_US"); if (<warning descr="Compare collation results explicitly.">$c->compare("cobalt","cobalt") === true</warning>) { echo "equal"; }
