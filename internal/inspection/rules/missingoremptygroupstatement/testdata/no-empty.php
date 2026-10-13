<?php
foreach ($queue as $job) {
    // nothing yet
}
while (poll()) {}
<weak_warning descr="Use a braced block for the body of this construct.">while</weak_warning> (poll()) tick();
