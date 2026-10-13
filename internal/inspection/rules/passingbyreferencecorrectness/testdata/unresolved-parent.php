<?php
// A recovered parent call outside a class has no resolvable contract.
function unresolvedParent() {
    parent::consume(1);
}
