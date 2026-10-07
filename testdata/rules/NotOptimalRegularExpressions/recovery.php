<?php
// Empty array elements do not compile in PHP; they are skipped.
preg_replace(['/a/', , '/b/'], 'x', $text);
