<?php
function g($j) {
	if ($j) {
		echo 1;
	}
        for ($i = 0; $i < 3; $i++) {
            <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($i == 0) echo "a";
        }
}
