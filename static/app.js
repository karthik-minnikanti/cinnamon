// Theme handling
const themeToggle = document.querySelector('.theme-toggle');
const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');

function setTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    themeToggle.innerHTML = theme === 'dark' ? '<i class="fas fa-sun"></i>' : '<i class="fas fa-moon"></i>';
}

// Initialize theme
const savedTheme = localStorage.getItem('theme') || (prefersDarkScheme.matches ? 'dark' : 'light');
setTheme(savedTheme);

themeToggle.addEventListener('click', () => {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    setTheme(currentTheme === 'dark' ? 'light' : 'dark');
});

// Navigation
const navLinks = document.querySelectorAll('.nav-links li');
const views = document.querySelectorAll('.view');

navLinks.forEach(link => {
    link.addEventListener('click', () => {
        const viewId = link.getAttribute('data-view');

        // Update active states
        navLinks.forEach(l => l.classList.remove('active'));
        link.classList.add('active');

        // Show selected view
        views.forEach(view => {
            view.classList.remove('active');
            if (view.id === viewId) {
                view.classList.add('active');
            }
        });
    });
});

// Charts initialization
const charts = {
    connectionTrends: null,
    serviceDistribution: null,
    connectionVolume: null,
    serviceTypes: null,
    errorDistribution: null,
    latencyTrends: null,
    errorTimeline: null
};

function initializeCharts() {
    // Connection Trends Chart
    charts.connectionTrends = new Chart(document.getElementById('connection-trends'), {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Connections',
                data: [],
                borderColor: '#000000',
                tension: 0.4,
                fill: false
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(0, 0, 0, 0.1)'
                    }
                },
                x: {
                    grid: {
                        display: false
                    }
                }
            }
        }
    });

    // Service Distribution Chart
    charts.serviceDistribution = new Chart(document.getElementById('service-distribution'), {
        type: 'doughnut',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    '#000000',
                    '#333333',
                    '#666666',
                    '#999999',
                    '#cccccc'
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                }
            }
        }
    });

    // Connection Volume Chart
    charts.connectionVolume = new Chart(document.getElementById('connection-volume'), {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Volume',
                data: [],
                borderColor: '#000000',
                tension: 0.4,
                fill: true,
                backgroundColor: 'rgba(0, 0, 0, 0.1)'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(0, 0, 0, 0.1)'
                    }
                },
                x: {
                    grid: {
                        display: false
                    }
                }
            }
        }
    });

    // Service Types Chart
    charts.serviceTypes = new Chart(document.getElementById('service-types'), {
        type: 'bar',
        data: {
            labels: [],
            datasets: [{
                label: 'Count',
                data: [],
                backgroundColor: '#000000'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(0, 0, 0, 0.1)'
                    }
                },
                x: {
                    grid: {
                        display: false
                    }
                }
            }
        }
    });

    // Error Distribution Chart
    charts.errorDistribution = new Chart(document.getElementById('error-distribution'), {
        type: 'pie',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    '#ff3d00',
                    '#ff6d00',
                    '#ff9100',
                    '#ffb300',
                    '#ffd600'
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                }
            }
        }
    });

    // Latency Trends Chart
    charts.latencyTrends = new Chart(document.getElementById('latency-trends'), {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Latency',
                data: [],
                borderColor: '#000000',
                tension: 0.4,
                fill: false
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(0, 0, 0, 0.1)'
                    }
                },
                x: {
                    grid: {
                        display: false
                    }
                }
            }
        }
    });

    // Error Timeline Chart
    charts.errorTimeline = new Chart(document.getElementById('error-timeline-chart'), {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Errors',
                data: [],
                borderColor: '#ff3d00',
                tension: 0.4,
                fill: true,
                backgroundColor: 'rgba(255, 61, 0, 0.1)'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(0, 0, 0, 0.1)'
                    }
                },
                x: {
                    grid: {
                        display: false
                    }
                }
            }
        }
    });
}

// Data fetching and updating
async function fetchData() {
    try {
        const [statsResponse, connectionsResponse] = await Promise.all([
            fetch('/api/connections/stats'),
            fetch('/api/connections')
        ]);

        const stats = await statsResponse.json();
        const connections = await connectionsResponse.json();

        updateDashboard(stats);
        updateConnections(connections.connections);
        updateCharts(stats);
    } catch (error) {
        console.error('Error fetching data:', error);
    }
}

function updateDashboard(stats) {
    // Update stat cards
    document.getElementById('total-connections').textContent = stats.total_connections || 0;
    document.getElementById('active-services').textContent = Object.keys(stats.service_type_stats || {}).length;
    document.getElementById('error-rate').textContent = calculateErrorRate(stats) + '%';
    document.getElementById('avg-latency').textContent = Math.round(stats.avg_latency || 0) + 'ms';

    // Update error stats
    document.getElementById('total-errors').textContent = calculateTotalErrors(stats);
    document.getElementById('error-percentage').textContent = calculateErrorRate(stats) + '%';
    document.getElementById('common-error').textContent = getMostCommonError(stats);
}

function updateConnections(connections) {
    const tbody = document.getElementById('recent-connections-body');
    tbody.innerHTML = '';

    // Filter for connections with errors
    const errorConnections = connections.filter(conn => conn.error);

    errorConnections.forEach(conn => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${conn.service_name || '-'}</td>
            <td>${conn.source_ip}:${conn.source_port}</td>
            <td>${conn.dest_ip}:${conn.dest_port}</td>
            <td>
                <span class="status-badge error">
                    ${conn.error}
                </span>
            </td>
            <td>${new Date(conn.timestamp).toLocaleString()}</td>
        `;
        tbody.appendChild(row);
    });

    // Update filters
    updateFilters(connections);
}

function updateCharts(stats) {
    // Update connection trends
    const timeLabels = generateTimeLabels(24);
    charts.connectionTrends.data.labels = timeLabels;
    charts.connectionTrends.data.datasets[0].data = generateRandomData(24);
    charts.connectionTrends.update();

    // Update service distribution
    const serviceData = Object.entries(stats.service_type_stats || {});
    charts.serviceDistribution.data.labels = serviceData.map(([type]) => type);
    charts.serviceDistribution.data.datasets[0].data = serviceData.map(([, count]) => count);
    charts.serviceDistribution.update();

    // Update other charts similarly...
}

// Helper functions
function calculateErrorRate(stats) {
    const totalErrors = calculateTotalErrors(stats);
    const totalConnections = stats.total_connections || 0;
    // Only count connections that have errors in the numerator
    const connectionsWithErrors = Object.keys(stats.error_counts || {}).length;
    return totalConnections > 0 ? ((connectionsWithErrors / totalConnections) * 100).toFixed(1) : '0.0';
}

function calculateTotalErrors(stats) {
    return Object.values(stats.error_counts || {}).reduce((sum, count) => sum + count, 0);
}

function getMostCommonError(stats) {
    const errors = Object.entries(stats.error_counts || {});
    if (errors.length === 0) return 'None';
    return errors.reduce((a, b) => a[1] > b[1] ? a : b)[0];
}

function generateTimeLabels(hours) {
    const labels = [];
    for (let i = hours - 1; i >= 0; i--) {
        const date = new Date();
        date.setHours(date.getHours() - i);
        labels.push(date.toLocaleTimeString());
    }
    return labels;
}

function generateRandomData(points) {
    return Array.from({ length: points }, () => Math.floor(Math.random() * 100));
}

function updateFilters(connections) {
    const services = [...new Set(connections.map(c => c.service_name).filter(Boolean))];
    const errors = [...new Set(connections.map(c => c.error).filter(Boolean))];
    const environments = [...new Set(connections.map(c => c.environment).filter(Boolean))];

    updateFilterOptions('service-filter', services);
    updateFilterOptions('error-filter', errors);
    updateFilterOptions('environment-filter', environments);
}

function updateFilterOptions(filterId, options) {
    const select = document.getElementById(filterId);
    select.innerHTML = '<option value="">All</option>';
    options.forEach(option => {
        const opt = document.createElement('option');
        opt.value = option;
        opt.textContent = option;
        select.appendChild(opt);
    });
}

// Settings handling
document.getElementById('refresh-interval').addEventListener('change', (e) => {
    const interval = parseInt(e.target.value) * 1000;
    clearInterval(window.refreshInterval);
    window.refreshInterval = setInterval(fetchData, interval);
});

document.getElementById('theme-select').addEventListener('change', (e) => {
    setTheme(e.target.value);
});

document.getElementById('notifications-toggle').addEventListener('change', (e) => {
    if (e.target.checked) {
        requestNotificationPermission();
    }
});

async function requestNotificationPermission() {
    try {
        const permission = await Notification.requestPermission();
        if (permission !== 'granted') {
            document.getElementById('notifications-toggle').checked = false;
        }
    } catch (error) {
        console.error('Error requesting notification permission:', error);
        document.getElementById('notifications-toggle').checked = false;
    }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initializeCharts();
    fetchData();
    window.refreshInterval = setInterval(fetchData, 5000);
}); 