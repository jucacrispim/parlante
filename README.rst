Parlante: web comments powered by sqlite
========================================

Parlante is a simple comments system written in go that uses sqlite
as a database.


Installation
------------

To install use:

.. code-block:: sh

   $ go install github.com/jucacrispim/parlante/cmd/parlante
   $ go install github.com/jucacrispim/parlante/cmd/parlante-tui


Usage
-----

To use parlante, first you need to add a new client a domain to the client.
Add them using the parlante-tui program:

.. code-block:: sh

   $ parlante-tui -dbpath /path/to/my/sqlite.db


Now start the server:

.. code-block:: sh

   $ parlante -dbpath /path/to/my/sqlite.db


To display the comments in your web page use this:

.. code-block:: html

   <div id="#comments-container"></div>
   <script src="<PARLANTE_URL>/parlante.js" async onload="parlanteLoadComments('<PARLANTE_URL>', '<CLIENT_UUID>', 'comments-container')"></script>


For more information check `the full documentation <https://docs.poraodojuca.dev/parlante/>`_
