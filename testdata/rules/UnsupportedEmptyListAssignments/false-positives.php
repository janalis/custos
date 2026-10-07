<?php
foreach ($queue as list(, $second)) {
    use_it($second);
}
foreach ($queue as [$first]) {
    use_it($first);
}
foreach ($queue as list($this->slot)) {
}
list(, $only) = $queue;
