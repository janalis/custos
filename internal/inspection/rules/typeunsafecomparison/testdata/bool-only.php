<?php
function flag(?bool $b) {
    return <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$b == null</weak_warning>;
}
