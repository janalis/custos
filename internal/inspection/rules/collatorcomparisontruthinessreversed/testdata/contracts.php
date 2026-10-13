<?php
$c=new Collator("en_US"); if (<warning descr="Compare collation results explicitly.">$c->compare("cobalt","cobalt") === true</warning>) { echo "equal"; }
