<?php
if (filter_var('false', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === null) { throw new InvalidArgumentException(); }
