<?php
function f($c) { $v=pg_escape_literal($c,'Bex'); pg_query($c,<warning descr="Use the already quoted PostgreSQL literal directly.">"SELECT '$v'"</warning>); }
