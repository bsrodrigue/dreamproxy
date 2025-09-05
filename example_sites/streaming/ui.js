document.addEventListener('DOMContentLoaded', () => {
  let isPlaying = false;
  let currentProgress = 35;

  document.getElementById('playButton').addEventListener('click', function () {
    togglePlay();
  });

  document.getElementById('watchButton').addEventListener('click', function () {
    togglePlay();
  });

  function togglePlay() {
    const playButton = document.getElementById('playButton');
    const loadingIndicator = document.getElementById('loadingIndicator');

    if (!isPlaying) {
      // Show loading
      playButton.style.display = 'none';
      loadingIndicator.style.display = 'block';

      // Simulate loading time
      setTimeout(() => {
        loadingIndicator.style.display = 'none';
        isPlaying = true;
        startProgressAnimation();
        showNotification('Now playing: Episode 01');
      }, 2000);
    } else {
      playButton.style.display = 'flex';
      isPlaying = false;
      stopProgressAnimation();
      showNotification('Playback paused');
    }
  }

  let progressInterval;

  function startProgressAnimation() {
    progressInterval = setInterval(() => {
      if (currentProgress < 100) {
        currentProgress += 0.1;
        document.getElementById('progressFill').style.width = currentProgress + '%';
        updateTime();
      }
    }, 100);
  }

  function stopProgressAnimation() {
    clearInterval(progressInterval);
  }

  function updateTime() {
    const totalSeconds = Math.floor((currentProgress / 100) * (42 * 60 + 18));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    document.getElementById('currentTime').textContent =
      minutes + ':' + seconds.toString().padStart(2, '0');
  }

  function selectEpisode(episodeNumber) {
    // Remove active class from all episodes
    document.querySelectorAll('.episode-item').forEach(item => {
      item.classList.remove('active');
    });

    // Add active class to selected episode
    event.currentTarget.classList.add('active');

    showNotification(`Selected Episode ${episodeNumber.toString().padStart(2, '0')}`);
  }

  function addToList() {
    showNotification('Added to your watchlist!');
  }

  function shareEpisode() {
    showNotification('Share link copied to clipboard!');
  }

  function showNotification(message) {
    // Create notification element
    const notification = document.createElement('div');
    notification.style.cssText = `
                position: fixed;
                top: 100px;
                right: 20px;
                background: linear-gradient(45deg, #4ecdc4, #44a08d);
                color: white;
                padding: 1rem 2rem;
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
                z-index: 1000;
                transform: translateX(100%);
                transition: transform 0.3s ease;
            `;
    notification.textContent = message;

    document.body.appendChild(notification);

    // Animate in
    setTimeout(() => {
      notification.style.transform = 'translateX(0)';
    }, 100);

    // Animate out and remove
    setTimeout(() => {
      notification.style.transform = 'translateX(100%)';
      setTimeout(() => {
        document.body.removeChild(notification);
      }, 300);
    }, 3000);
  }

  // Keyboard shortcuts
  document.addEventListener('keydown', function (e) {
    if (e.code === 'Space') {
      e.preventDefault();
      togglePlay();
    } else if (e.code === 'ArrowLeft') {
      e.preventDefault();
      currentProgress = Math.max(0, currentProgress - 5);
      document.getElementById('progressFill').style.width = currentProgress + '%';
      updateTime();
    } else if (e.code === 'ArrowRight') {
      e.preventDefault();
      currentProgress = Math.min(100, currentProgress + 5);
      document.getElementById('progressFill').style.width = currentProgress + '%';
      updateTime();
    }
  });

  // Initialize
  updateTime();

  // Welcome message
  setTimeout(() => {
    showNotification('Welcome to StreamFlix! Press Space to play/pause.');
  }, 1000);
});
