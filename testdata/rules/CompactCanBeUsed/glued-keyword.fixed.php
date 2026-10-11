<?php
function glued($a, $b) { return compact('a', 'b'); }
function shown($a, $b) { echo compact('a', 'b') === [] ? 0 : 1; }
