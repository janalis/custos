<?php
libxml_use_internal_errors(true); $d=new DOMDocument(); <warning descr="Clear collected XML errors during repeated parsing.">foreach($docs as $xml){$d->loadXML($xml);}</warning>
