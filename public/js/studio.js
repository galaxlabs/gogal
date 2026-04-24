(function () {
  function slugify(value) {
    return String(value || '').trim().toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '');
  }

  function setupBuilder() {
    var root = document.querySelector('.builder-page');
    if (!root || root.dataset.bound === '1') return;
    root.dataset.bound = '1';

    var state = { fields: [] };
    var rows = root.querySelector('#field-rows');
    var preview = root.querySelector('#doctype-preview');
    var result = root.querySelector('#doctype-save-result');

    var nameInput = root.querySelector('#dt-name');
    var labelInput = root.querySelector('#dt-label');
    var moduleInput = root.querySelector('#dt-module');
    var tableInput = root.querySelector('#dt-table');
    var descriptionInput = root.querySelector('#dt-description');
    var singleInput = root.querySelector('#dt-single');
    var childInput = root.querySelector('#dt-child');

    var isEditMode = false;
    try {
      var existing = JSON.parse(root.dataset.doctypeJson || '{}');
      if (existing && existing.doctype) {
        isEditMode = true;
        nameInput.value = existing.doctype || '';
        labelInput.value = existing.label || '';
        moduleInput.value = existing.module || 'Core';
        tableInput.value = existing.table_name || '';
        descriptionInput.value = existing.description || '';
        singleInput.checked = !!existing.is_single;
        childInput.checked = !!existing.is_child_table;
        state.fields = (existing.fields || []).map(function (f, i) {
          return {
            id: String(Date.now() + i),
            fieldname: f.fieldname || '',
            label: f.label || '',
            fieldtype: f.fieldtype || 'Data',
            options: f.options || '',
            reqd: !!f.reqd,
            unique: !!f.unique
          };
        });
      }
    } catch (e) {}

    function buildPayload() {
      return {
        doctype: nameInput.value.trim(),
        label: labelInput.value.trim() || nameInput.value.trim(),
        module: moduleInput.value.trim() || 'Core',
        table_name: tableInput.value.trim(),
        description: descriptionInput.value.trim(),
        is_single: !!singleInput.checked,
        is_child_table: !!childInput.checked,
        fields: state.fields.map(function (f, idx) {
          return {
            fieldname: slugify(f.fieldname || f.label),
            label: f.label || f.fieldname,
            fieldtype: f.fieldtype || 'Data',
            options: f.options || '',
            reqd: !!f.reqd,
            unique: !!f.unique,
            sort_order: idx + 1
          };
        })
      };
    }

    function renderPreview() {
      preview.textContent = JSON.stringify(buildPayload(), null, 2);
    }

    function bindRowEvents() {
      rows.querySelectorAll('.field-row').forEach(function (row) {
        var id = row.dataset.id;
        var field = state.fields.find(function (f) { return f.id === id; });
        if (!field) return;

        row.querySelectorAll('[data-k]').forEach(function (input) {
          var key = input.dataset.k;
          var handler = function () {
            field[key] = input.type === 'checkbox' ? input.checked : input.value;
            if (key === 'label' && !field.fieldname) {
              field.fieldname = slugify(field.label);
              var nameInputLocal = row.querySelector('[data-k="fieldname"]');
              if (nameInputLocal) nameInputLocal.value = field.fieldname;
            }
            renderPreview();
          };
          input.addEventListener('input', handler);
          input.addEventListener('change', handler);
        });

        var rm = row.querySelector('[data-remove]');
        if (rm) {
          rm.addEventListener('click', function () {
            state.fields = state.fields.filter(function (f) { return f.id !== id; });
            renderRows();
          });
        }
      });
    }

    function renderRows() {
      rows.innerHTML = '';
      if (!state.fields.length) {
        rows.innerHTML = '<div class="muted">No fields yet. Click Add Field.</div>';
      }
      state.fields.forEach(function (f) {
        var row = document.createElement('div');
        row.className = 'field-row';
        row.dataset.id = f.id;
        row.innerHTML = '' +
          '<input data-k="label" value="' + (f.label || '') + '" placeholder="Label" />' +
          '<input data-k="fieldname" value="' + (f.fieldname || '') + '" placeholder="field_name" />' +
          '<select data-k="fieldtype">' +
            ['Data','Text','Int','Float','Currency','Date','Datetime','Check','Select','Link','Table'].map(function (v) {
              return '<option value="' + v + '" ' + (f.fieldtype===v?'selected':'') + '>' + v + '</option>';
            }).join('') +
          '</select>' +
          '<button class="btn" data-remove="1" type="button">Remove</button>' +
          '<input data-k="options" value="' + (f.options || '') + '" placeholder="Options / target DocType" style="grid-column:1 / span 2" />' +
          '<label><input type="checkbox" data-k="reqd" ' + (f.reqd ? 'checked' : '') + ' /> Required</label>' +
          '<label><input type="checkbox" data-k="unique" ' + (f.unique ? 'checked' : '') + ' /> Unique</label>' +
          '<span class="drag">☰</span>';
        rows.appendChild(row);
      });
      bindRowEvents();
      renderPreview();

      if (window.Sortable && rows._sortableBound !== true) {
        rows._sortableBound = true;
        new Sortable(rows, {
          handle: '.drag',
          animation: 120,
          onEnd: function () {
            var newOrder = [];
            rows.querySelectorAll('.field-row').forEach(function (node) {
              var id = node.dataset.id;
              var found = state.fields.find(function (f) { return f.id === id; });
              if (found) newOrder.push(found);
            });
            state.fields = newOrder;
            renderPreview();
          }
        });
      }
    }

    root.querySelector('#add-field-btn').addEventListener('click', function () {
      state.fields.push({ id: String(Date.now() + Math.random()), fieldname: '', label: 'New Field', fieldtype: 'Data', options: '', reqd: false, unique: false });
      renderRows();
    });

    root.querySelector('#doctype-save-btn').addEventListener('click', function () {
      var payload = buildPayload();
      var url = isEditMode ? '/api/doctypes/' + encodeURIComponent(payload.doctype) : '/api/doctypes';
      fetch(url, {
        method: isEditMode ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      }).then(function (r) {
        return r.json().then(function (body) {
          return { ok: r.ok, body: body };
        });
      }).then(function (res) {
        if (!res.ok) {
          result.innerHTML = '<div class="alert alert-error">Save failed: ' + (res.body.error && res.body.error.message ? res.body.error.message : JSON.stringify(res.body)) + '</div>';
          return;
        }
        var name = payload.doctype;
        result.innerHTML = '<div class="alert alert-success">DocType ' + (isEditMode ? 'updated' : 'saved') + '. ' +
          '<a hx-get="/desk/resource/' + name + '" hx-target="#app-content" hx-swap="innerHTML">Open List</a> | ' +
          '<a hx-get="/desk/resource/' + name + '/new" hx-target="#app-content" hx-swap="innerHTML">Open Form</a></div>';
      }).catch(function (e) {
        result.innerHTML = '<div class="alert alert-error">Save failed: ' + e + '</div>';
      });
    });

    renderRows();
  }

  function setupRecordForm() {
    var root = document.querySelector('.record-form-page');
    if (!root || root.dataset.bound === '1') return;
    root.dataset.bound = '1';

    var saveBtn = root.querySelector('#record-save-btn');
    var result = root.querySelector('#record-save-result');

    function collectPayload() {
      var payload = {};
      root.querySelectorAll('[data-input]').forEach(function (input) {
        var key = input.dataset.input;
        payload[key] = input.type === 'checkbox' ? input.checked : input.value;
      });
      return payload;
    }

    saveBtn.addEventListener('click', function () {
      var doctype = root.dataset.doctype;
      var recordID = root.dataset.recordId;
      var isNew = root.dataset.isNew === 'true';
      var url = '/api/resource/' + doctype + (isNew ? '' : '/' + encodeURIComponent(recordID));

      fetch(url, {
        method: isNew ? 'POST' : 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(collectPayload())
      }).then(function (r) {
        return r.json().then(function (body) {
          return { ok: r.ok, body: body };
        });
      }).then(function (res) {
        if (!res.ok) {
          result.innerHTML = '<div class="alert alert-error">Save failed: ' + (res.body.error && res.body.error.message ? res.body.error.message : JSON.stringify(res.body)) + '</div>';
          return;
        }
        result.innerHTML = '<div class="alert alert-success">Record saved. <a hx-get="/desk/resource/' + doctype + '" hx-target="#app-content" hx-swap="innerHTML">Back to list</a></div>';
      }).catch(function (e) {
        result.innerHTML = '<div class="alert alert-error">Save failed: ' + e + '</div>';
      });
    });

    root.querySelectorAll('input[data-link-target]').forEach(function (input) {
      var listId = 'link-options-' + input.dataset.input;
      var datalist = document.createElement('datalist');
      datalist.id = listId;
      input.setAttribute('list', listId);
      input.parentNode.appendChild(datalist);

      var timeout = null;
      input.addEventListener('input', function () {
        clearTimeout(timeout);
        timeout = setTimeout(function () {
          var target = input.dataset.linkTarget;
          if (!target) return;
          fetch('/api/resource/' + encodeURIComponent(target) + '/link-search?search=' + encodeURIComponent(input.value || '') + '&limit=8')
            .then(function (r) { return r.json(); })
            .then(function (body) {
              datalist.innerHTML = '';
              (body.data || []).forEach(function (opt) {
                var o = document.createElement('option');
                o.value = opt.name;
                datalist.appendChild(o);
              });
            }).catch(function () {});
        }, 220);
      });
    });
  }

  function boot() {
    setupBuilder();
    setupRecordForm();
  }

  document.addEventListener('DOMContentLoaded', boot);
  document.body.addEventListener('htmx:afterSwap', boot);
})();
