<?php
foreach ($queue as $job) {
    // nothing yet
}
while (poll()) {}
while (poll()) {
    tick();
}
