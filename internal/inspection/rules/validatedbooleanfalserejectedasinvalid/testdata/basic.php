<?php
if (<warning descr="Distinguish valid false from invalid boolean input.">filter_var('false', FILTER_VALIDATE_BOOLEAN) === false</warning>) { throw new InvalidArgumentException(); }
