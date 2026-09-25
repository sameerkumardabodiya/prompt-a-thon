function showSection(id) {
    document.querySelectorAll('section').forEach(function(s) { s.classList.remove('active'); });
    document.querySelectorAll('nav a').forEach(function(a) { a.classList.remove('active'); });
    var targetSec = document.getElementById('sec-' + id);
    var targetNav = document.getElementById('nav-' + id);
    if (targetSec) targetSec.classList.add('active');
    if (targetNav) targetNav.classList.add('active');
}

var sessionEvents = [];
var previousNodeStates = {};

function logEvent(msg) {
    var time = new Date().toLocaleTimeString();
    sessionEvents.unshift({ time: time, msg: msg });
    renderEvents();
}

function renderEvents() {
    var container = document.getElementById('events-timeline');
    if (!container) return;
    if (sessionEvents.length === 0) {
        container.innerHTML = '<div style="color:var(--text-secondary); font-size:14px;">Awaiting telemetry...</div>';
        return;
    }
    var html = '';
    for (var i = 0; i < sessionEvents.length; i++) {
        html += '<div class="event-row">' +
                '<div class="event-time">' + sessionEvents[i].time + '</div>' +
                '<div class="event-msg">' + sessionEvents[i].msg + '</div>' +
                '</div>';
    }
    container.innerHTML = html;
}

async function fetchStatus() {
    try {
        var res = await fetch('/api/status');
        var data = await res.json();
        
        var clusterHTML = '';
        var onlineCount = 0;

        data.nodes.forEach(function(n, i) {
            var online = n.status === 'ONLINE';
            if (online) onlineCount++;
            var nodeName = 'NODE-0' + (i + 1);
            
            if (previousNodeStates[nodeName] !== undefined && previousNodeStates[nodeName] !== online) {
                logEvent('State Transition: ' + nodeName + ' is now ' + (online ? 'ONLINE' : 'OFFLINE'));
            }
            previousNodeStates[nodeName] = online;

            var topEl = document.getElementById('top-node-' + (i + 1));
            var badgeEl = document.getElementById('badge-n' + (i + 1));
            if (topEl && badgeEl) {
                topEl.className = 'top-node ' + (online ? 'online' : 'offline');
                badgeEl.className = 'badge';
                badgeEl.style.backgroundColor = online ? 'var(--online-bg)' : 'var(--offline-bg)';
                badgeEl.style.color = online ? 'var(--online)' : 'var(--offline)';
                badgeEl.style.borderColor = 'transparent';
                badgeEl.innerText = online ? 'SYNCED' : 'HALTED';
            }

            clusterHTML += '<div class="surface" style="padding:24px;">' +
                '<div class="surface-title" style="margin-bottom:16px;">' +
                    '<span style="font-family:var(--font-mono);">' + nodeName + '</span>' +
                    '<span class="badge" style="background:' + (online ? 'var(--online-bg)' : 'var(--offline-bg)') + '; color:' + (online ? 'var(--online)' : 'var(--offline)') + ';">' + n.status + '</span>' +
                '</div>' +
                '<div style="display:flex; gap:40px;">' +
                    '<div><div class="metric-label">Instance Route</div><div style="font-family:var(--font-mono); font-size:13px;">' + n.url + '</div></div>' +
                    '<div><div class="metric-label">Latency / IOPS</div><div class="badge badge-muted">Backend Not Emitting</div></div>' +
                    '<div><div class="metric-label">Storage Capacity</div><div class="badge badge-muted">Backend Not Emitting</div></div>' +
                '</div>' +
            '</div>';
        });

        var countEl = document.getElementById('stat-nodes-count');
        if (countEl) countEl.innerText = onlineCount + '/' + data.nodes.length;

        var clList = document.getElementById('cluster-list');
        if (clList) clList.innerHTML = clusterHTML;
    } catch (e) {
        console.error('Failed to fetch status');
    }
}

async function fetchFiles() {
    try {
        var res = await fetch('/api/files');
        var data = await res.json();
        var tbody = document.getElementById('objects-table-body');
        var statFiles = document.getElementById('stat-objects-count');
        
        if (statFiles) statFiles.innerText = data.files.length;
        if (!tbody) return;

        if (data.files.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; color:var(--text-secondary); padding:40px;">Namespace is currently empty.</td></tr>';
            return;
        }

        var html = '';
        data.files.forEach(function(f) {
            html += '<tr>' +
                    '<td style="font-weight:500;">' + f + '</td>' +
                    '<td><span class="badge badge-muted">N/A</span></td>' +
                    '<td><span class="badge badge-muted">N/A</span></td>' +
                    '<td><span class="badge badge-muted">NOT AVAILABLE</span></td>' +
                    '<td><button class="btn btn-ghost" style="padding:6px 12px; margin:0;" onclick="retrieveObject(\'' + f + '\')">Retrieve</button></td>' +
                    '</tr>';
        });
        tbody.innerHTML = html;
    } catch (e) {
        console.error('Failed to fetch files');
    }
}

async function uploadFile(input) {
    var file = input.files[0];
    if (!file) return;

    logEvent('Ingest initialized for object payload: ' + file.name);
    
    try {
        var res = await fetch('/objects/' + file.name, { method: 'PUT', body: file });
        if (res.ok) {
            logEvent('Commit successful. Payload replicated: ' + file.name);
            fetchFiles();
        } else {
            logEvent('Cluster rejected payload: ' + file.name);
        }
    } catch (e) {
        logEvent('Network timeout during ingress of: ' + file.name);
    }
    input.value = '';
}

function retrieveObject(filename) {
    logEvent('Router fetching healthiest replica for: ' + filename);
    window.open('/objects/' + filename, '_blank');
}

setInterval(fetchStatus, 2000);
window.onload = function() {
    fetchStatus();
    fetchFiles();
    logEvent("Vault Core loaded. Gateway accepting connections.");
};
