<?php
function both(<warning descr="Objects are handed over by handle already; drop the '&' before '$c'.">Countable&Traversable &$c</warning>) {}
function dnf(<warning descr="Objects are handed over by handle already; drop the '&' before '$x'.">(Countable&Traversable)|null &$x</warning>) {}
