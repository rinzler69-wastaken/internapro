<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import { auth } from '../lib/stores';
  import LoadingSpinner from '../components/LoadingSpinner.svelte';
  import Chart from 'chart.js/auto';

  let loading = true;
  let weeklyTrends = null;
  let patterns = null;
  let insights = null;
  let internId = null;

  let weeklyChart;
  let patternsChart;

  onMount(async () => {
    // TODO: Get intern ID from user data when intern handler is implemented
    // For now, using mock ID
    internId = 1;

    try {
      const [trendsRes, patternsRes, insightsRes] = await Promise.all([
        api.getWeeklyTrends(internId),
        api.getCheckInPatterns(internId, 30),
        api.getPerformanceInsights(internId)
      ]);

      weeklyTrends = trendsRes.data;
      patterns = patternsRes.data;
      insights = insightsRes.data;

      // Render charts after data is loaded
      setTimeout(() => {
        renderWeeklyChart();
        renderPatternsChart();
      }, 100);

    } catch (error) {
      console.error('Failed to load analytics:', error);
    } finally {
      loading = false;
    }
  });

  function renderWeeklyChart() {
    const ctx = document.getElementById('weeklyChart');
    if (!ctx || !weeklyTrends) return;

    const labels = weeklyTrends.daily_records.map(r => r.day_of_week.substring(0, 3));
    const times = weeklyTrends.daily_records.map(r => {
      if (!r.check_in_hour) return null;
      return r.check_in_hour * 60 + r.check_in_minute;
    });

    weeklyChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels,
        datasets: [{
          label: 'Check-in Time',
          data: times,
          borderColor: '#000',
          backgroundColor: 'rgba(0, 0, 0, 0.05)',
          tension: 0.3,
          fill: true,
          pointRadius: 4,
          pointHoverRadius: 6,
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
            beginAtZero: false,
            min: 420, // 7:00 AM
            max: 600, // 10:00 AM
            ticks: {
              callback: function(value) {
                const hours = Math.floor(value / 60);
                const mins = value % 60;
                return `${hours}:${mins.toString().padStart(2, '0')}`;
              }
            }
          }
        }
      }
    });
  }

  function renderPatternsChart() {
    const ctx = document.getElementById('patternsChart');
    if (!ctx || !patterns) return;

    // Filter to show only relevant hours (6 AM to 12 PM)
    const relevantPatterns = patterns.patterns.filter(p => p.hour >= 6 && p.hour <= 12);

    patternsChart = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: relevantPatterns.map(p => `${p.hour}:00`),
        datasets: [{
          label: 'Check-ins',
          data: relevantPatterns.map(p => p.count),
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: '#000',
          borderWidth: 1
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
            ticks: {
              stepSize: 1
            }
          }
        }
      }
    });
  }

  function getTendencyColor(tendency) {
    const colors = {
      'early_bird': 'badge-info',
      'on_time': 'badge-success',
      'frequently_late': 'badge-error',
      'inconsistent': 'badge-warning'
    };
    return colors[tendency] || 'badge-neutral';
  }

  function getTendencyText(tendency) {
    return tendency.replace(/_/g, ' ').toUpperCase();
  }

  function getScoreColor(score) {
    if (score >= 85) return 'text-green-600';
    if (score >= 70) return 'text-blue-600';
    if (score >= 50) return 'text-yellow-600';
    return 'text-red-600';
  }
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-3xl font-geist font-bold text-black">Performance Analytics</h1>
    <p class="mt-1 text-sm font-inter text-vercel-gray-600">
      Track your attendance trends and performance insights
    </p>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>
  {:else if weeklyTrends && patterns && insights}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="card p-6">
        <p class="text-sm font-inter text-vercel-gray-600 mb-1">Attendance Rate</p>
        <p class="text-3xl font-geist font-semibold text-black">
          {weeklyTrends.summary.attendance_rate.toFixed(1)}%
        </p>
        <div class="w-full bg-vercel-gray-200 rounded-full h-2 mt-2">
          <div 
            class="bg-black h-2 rounded-full transition-all duration-500" 
            style="width: {weeklyTrends.summary.attendance_rate}%"
          />
        </div>
      </div>

      <div class="card p-6">
        <p class="text-sm font-inter text-vercel-gray-600 mb-1">Punctuality Rate</p>
        <p class="text-3xl font-geist font-semibold text-black">
          {weeklyTrends.summary.punctuality_rate.toFixed(1)}%
        </p>
        <div class="w-full bg-vercel-gray-200 rounded-full h-2 mt-2">
          <div 
            class="bg-black h-2 rounded-full transition-all duration-500" 
            style="width: {weeklyTrends.summary.punctuality_rate}%"
          />
        </div>
      </div>

      <div class="card p-6">
        <p class="text-sm font-inter text-vercel-gray-600 mb-1">Avg. Check-in</p>
        <p class="text-3xl font-geist font-semibold text-black">
          {weeklyTrends.summary.average_check_in_time || '--'}
        </p>
      </div>

      <div class="card p-6">
        <p class="text-sm font-inter text-vercel-gray-600 mb-1">Tendency</p>
        <span class="badge {getTendencyColor(weeklyTrends.summary.tendency)} text-sm mt-2">
          {getTendencyText(weeklyTrends.summary.tendency)}
        </span>
      </div>
    </div>

    <!-- Charts -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Weekly Trends -->
      <div class="card p-6">
        <h2 class="text-xl font-geist font-semibold text-black mb-4">Weekly Check-in Trends</h2>
        <div style="height: 300px;">
          <canvas id="weeklyChart"></canvas>
        </div>
      </div>

      <!-- Patterns -->
      <div class="card p-6">
        <h2 class="text-xl font-geist font-semibold text-black mb-4">Check-in Patterns (30 Days)</h2>
        <div style="height: 300px;">
          <canvas id="patternsChart"></canvas>
        </div>
      </div>
    </div>

    <!-- Performance Insights -->
    <div class="card p-6">
      <h2 class="text-xl font-geist font-semibold text-black mb-4">💡 Performance Insights</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <!-- Strengths -->
        <div>
          <h3 class="font-geist font-semibold text-green-600 mb-3">💪 Strengths</h3>
          {#if insights.strengths && insights.strengths.length > 0}
            <ul class="space-y-2">
              {#each insights.strengths as strength}
                <li class="flex items-start space-x-2">
                  <svg class="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>
                  <span class="text-sm font-inter text-vercel-gray-700">{strength}</span>
                </li>
              {/each}
            </ul>
          {:else}
            <p class="text-sm font-inter text-vercel-gray-500">Keep working to build strengths!</p>
          {/if}
        </div>

        <!-- Concerns -->
        <div>
          <h3 class="font-geist font-semibold text-red-600 mb-3">⚠️ Areas for Improvement</h3>
          {#if insights.concerns && insights.concerns.length > 0}
            <ul class="space-y-2">
              {#each insights.concerns as concern}
                <li class="flex items-start space-x-2">
                  <svg class="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                  </svg>
                  <span class="text-sm font-inter text-vercel-gray-700">{concern}</span>
                </li>
              {/each}
            </ul>
          {:else}
            <p class="text-sm font-inter text-vercel-gray-500">No concerns identified!</p>
          {/if}
        </div>

        <!-- Suggestions -->
        <div>
          <h3 class="font-geist font-semibold text-blue-600 mb-3">💡 Suggestions</h3>
          {#if insights.suggestions && insights.suggestions.length > 0}
            <ul class="space-y-2">
              {#each insights.suggestions as suggestion}
                <li class="flex items-start space-x-2">
                  <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                  <span class="text-sm font-inter text-vercel-gray-700">{suggestion}</span>
                </li>
              {/each}
            </ul>
          {:else}
            <p class="text-sm font-inter text-vercel-gray-500">Keep up the great work!</p>
          {/if}
        </div>
      </div>

      <!-- Overall Score -->
      <div class="border-t border-vercel-gray-200 pt-6">
        <div class="text-center">
          <p class="text-sm font-inter text-vercel-gray-600 mb-2">Overall Performance Score</p>
          <p class="text-6xl font-geist font-bold {getScoreColor(insights.overall_score)}">
            {insights.overall_score}
          </p>
          <p class="text-sm font-inter text-vercel-gray-500 mt-1">out of 100</p>
        </div>
      </div>
    </div>
  {:else}
    <div class="card p-12 text-center">
      <p class="text-sm font-inter text-vercel-gray-600">
        No analytics data available yet. Complete your attendance to see insights.
      </p>
    </div>
  {/if}
</div>
