<?php
// @noinspection CollatorComparisonTruthinessReversed
$c=new Collator("en_US"); if ($c->compare("cobalt","cobalt") === true) { echo "equal"; }
